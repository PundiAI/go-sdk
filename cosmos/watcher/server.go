package watcher

import (
	"context"
	"crypto/sha256"
	"sync"
	"time"

	"github.com/cometbft/cometbft/libs/bytes"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/functionx/fx-core/v8/client/grpc"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/pundiai/go-sdk/cosmos"
	"github.com/pundiai/go-sdk/db"
	"github.com/pundiai/go-sdk/log"
	"github.com/pundiai/go-sdk/server"
)

type RPCClient interface {
	TxByHash(txHash string) (*types.TxResponse, error)
	GetLatestBlock() (*cmtservice.Block, error)
}

type Handler interface {
	Enabled() bool
	HandleBlock(ctx context.Context, block Block) error
}

var _ server.Server = (*Server)(nil)

type Server struct {
	logger log.Logger
	config Config

	codec    codec.Codec
	client   RPCClient
	handlers []Handler
}

func NewServer(logger log.Logger, config Config, codec codec.Codec, handlers ...Handler) *Server {
	return &Server{
		config:   config,
		logger:   logger.With("server", "cosmos-watcher"),
		codec:    codec,
		handlers: handlers,
	}
}

func (s *Server) RegisterHandler(handler Handler) {
	s.handlers = append(s.handlers, handler)
}

func (s *Server) Init(ctx context.Context, db db.DB) (err error) {
	if !s.config.Enabled {
		return nil
	}
	s.client, err = grpc.DailClient(s.config.GrpcUrl, ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Start(ctx context.Context, group *errgroup.Group) (err error) {
	if !s.config.Enabled {
		return nil
	}
	group.Go(func() error {
		return s.scanBlock(ctx)
	})
	return nil
}

func (s *Server) Close() error {
	if s.client == nil {
		return nil
	}
	s.logger.Info("close cosmos watcher")
	return nil
}

func (s *Server) scanBlock(ctx context.Context) error {
	if len(s.handlers) == 0 {
		s.logger.Warn("no handler registered")
	}
	startTime := time.Now()

	latestBlockHeight, err := s.getStartBlock()
	if err != nil {
		return err
	}
	s.logger.Info("start scan block", "startBlockHeight", s.config.StartBlockHeight, "endBlockHeight", s.config.EndBlockHeight)

	wg := sync.WaitGroup{}
	pools := make(chan struct{}, s.config.BatchHandler)
loop:
	for blockHeight := s.config.StartBlockHeight; s.config.EndBlockHeight < 0 || blockHeight < s.config.EndBlockHeight; {
		select {
		case <-ctx.Done():
			break loop
		default:
		}

		// wait for the latest block
		var newBlock Block
		if s.config.EndBlockHeight < 0 && blockHeight >= latestBlockHeight {
			time.Sleep(s.config.BlockInterval)

			latestBlock, err := s.fetchBlock()
			if err != nil {
				s.logger.Warnf("query latest block error: %s", err.Error())
				continue
			}

			latestBlockHeight = latestBlock.Height
			if latestBlockHeight <= blockHeight {
				continue
			}
			newBlock = latestBlock
		}
		if blockHeight%10 == 0 {
			s.logger.Info("new block", "height", blockHeight)
		}
		if blockHeight%10000 == 0 {
			s.logger.Infof("sync block rate: %f/s", 10000/time.Since(startTime).Seconds())
			startTime = time.Now()
		}

		pools <- struct{}{}
		wg.Add(1)
		go func(newBlock Block) {
			defer wg.Done()
			for _, h := range s.handlers {
				if !h.Enabled() {
					continue
				}
				if e := h.HandleBlock(ctx, newBlock); e != nil {
					s.logger.Warn("handle block failed", "error", e.Error(), "blockHeight", blockHeight)
					time.Sleep(100 * time.Millisecond)
					continue
				}
			}

			<-pools
		}(newBlock)

		blockHeight++
	}

	s.logger.Info("wait for all goroutines to finish")
	wg.Wait()

	return nil
}

func (s *Server) getStartBlock() (int64, error) {
	block, err := s.client.GetLatestBlock()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get latest block")
	}
	latestBlockHeight := block.Header.Height
	if s.config.StartBlockHeight <= -1 {
		s.config.StartBlockHeight = latestBlockHeight
	}
	if s.config.EndBlockHeight > 0 && s.config.StartBlockHeight > s.config.EndBlockHeight {
		return 0, errors.New("invalid block height params")
	}
	if s.config.EndBlockHeight == 0 {
		s.config.EndBlockHeight = latestBlockHeight
	}
	if s.config.StartBlockHeight == 0 {
		s.config.StartBlockHeight++
	}
	return latestBlockHeight, nil
}

func (s *Server) fetchBlock() (block Block, err error) {
	newBlock, err := s.client.GetLatestBlock()
	if err != nil {
		return block, errors.Wrap(err, "failed to get latest block")
	}

	txsData := newBlock.Data.Txs
	txsHash := make([]bytes.HexBytes, len(txsData))
	txs := make([]*tx.Tx, len(txsData))
	for i := 0; i < len(txsData); i++ {
		txs[i], err = cosmos.DecodeTx(s.codec, txsData[i])
		if err != nil {
			return block, errors.Wrap(err, "failed to decode tx")
		}
		txHash := sha256.Sum256(txsData[i])
		copy(txsHash[i], txHash[:])
	}

	block = Block{
		ChainID:   newBlock.Header.ChainID,
		Height:    newBlock.Header.Height,
		BlockTime: newBlock.Header.Time,
		TxsHash:   txsHash,
		Txs:       txs,
	}
	return block, nil
}
