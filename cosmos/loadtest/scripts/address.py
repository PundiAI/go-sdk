import bech32


class Address:
    def __init__(self, addr):
        if type(addr) == str:
            hrp, data = bech32.bech32_decode(addr)
            self._hrp = hrp
            self._data = data
        elif type(addr) == bytes:
            self._data = bech32.convertbits(addr, 8, 5)
        else:
            raise Exception('invalid address')

    def get_hrp(self) -> str:
        return self._hrp

    def to_string(self, hrp: str) -> str:
        if hrp is None:
            hrp = self._hrp
        return bech32.bech32_encode(hrp, self._data)

    def to_bytes(self) -> bytes:
        return bytes(bech32.convertbits(self._data, 5, 8, False))
