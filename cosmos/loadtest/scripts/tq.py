import json
import sys

import toml


def main():
    args = sys.argv[1:]

    if len(args) == 1:
        print(json.dumps(toml.load(args[0])))
        return
    if len(args) == 2:
        with open(args[0], 'w') as f:
            toml.dump(json.loads(args[1]), f)


if __name__ == "__main__":
    main()
