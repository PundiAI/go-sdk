import json
import os
import sys

import eth_keys

import address
import crypto

home_path = os.path.expanduser('~')


def add_key(address_prefix="cosmos"):
    mnemonic = crypto.generate_mnemonic()
    priv_key = crypto.mnemonic_to_privkey(mnemonic)
    if address_prefix == "0x":
        addr = eth_keys.keys.PrivateKey(priv_key.to_bytes()).public_key.to_checksum_address()
    else:
        addr = priv_key.to_address(address_prefix)
    return addr, mnemonic


def add_account(account_name, addr, mnemonic):
    test_accounts_dir = os.getenv("APP_TEST_ACCOUNTS_DIR", f"{home_path}/test_accounts")
    filename = f"{test_accounts_dir}/{account_name}.json"
    os.makedirs(os.path.dirname(filename), exist_ok=True)
    with open(filename, 'w') as f:
        data = {
            "address": addr,
            "mnemonic": mnemonic,
        }
        json.dump(data, f)


def create_genesis_account(account_index, default_account_length, address_prefix="cosmos", denom="stake", amount="1000000000000000000000000"):
    addr, mnemonic = add_key(address_prefix=address_prefix)
    account_name = f"test_{account_index - default_account_length}"
    add_account(account_name=account_name, addr=addr, mnemonic=mnemonic)
    return {
        "balance": {
            "address": addr,
            "coins": [
                {
                    "denom": denom,
                    "amount": amount
                }
            ]
        },
        "account": {
            "@type": "/cosmos.auth.v1beta1.BaseAccount",
            "address": addr,
            "pub_key": None,
            "account_number": f"{account_index}",
            "sequence": "0"
        }
    }


def read_genesis_file(genesis_json_file_path):
    print("Reading genesis file")
    with open(genesis_json_file_path, 'r') as f:
        return json.load(f)


def write_genesis_file(genesis_json_file_path, data):
    print("Writing results to genesis file")
    with open(genesis_json_file_path, 'w') as f:
        json.dump(data, f, indent=4)


def main():
    args = sys.argv[1:]
    number_of_accounts = int(args[0])
    if number_of_accounts < 1:
        return print("\nNumber of accounts must be greater than 0\n")
    genesis_file_path = args[1]

    genesis_file = read_genesis_file(genesis_file_path)

    default_account_length = len(genesis_file["app_state"]["auth"]["accounts"])
    admin_addr_str = genesis_file["app_state"]["auth"]["accounts"][0]["address"]
    # if address is starting with 0x, it is an ethereum address
    if admin_addr_str.startswith("0x"):
        address_prefix = "0x"
    else:
        admin_addr = address.Address(admin_addr_str)
        address_prefix = admin_addr.get_hrp()

    amount = genesis_file["app_state"]["bank"]["balances"][0]["coins"][0]["amount"]
    denom = genesis_file["app_state"]["bank"]["balances"][0]["coins"][0]["denom"]

    if len(genesis_file["app_state"]["bank"]["supply"]) > 0:
        supply_amount = genesis_file["app_state"]["bank"]["supply"][0]["amount"]
        genesis_file["app_state"]["bank"]["supply"][0][
            "amount"] = f"{number_of_accounts * int(amount) + int(supply_amount)}"

    global_accounts_mapping = {}
    for i in range(0, number_of_accounts):
        account_index = i + default_account_length
        global_accounts_mapping[account_index] = create_genesis_account(account_index, default_account_length, address_prefix, denom, amount)

    sorted_keys = sorted(list(global_accounts_mapping.keys()))
    account_info = [0] * len(sorted_keys)
    balances = [0] * len(sorted_keys)
    for key in sorted_keys:
        balances[key - default_account_length] = global_accounts_mapping[key]["balance"]
        account_info[key - default_account_length] = global_accounts_mapping[key]["account"]

    genesis_file["app_state"]["bank"]["balances"] = genesis_file["app_state"]["bank"]["balances"] + balances
    genesis_file["app_state"]["auth"]["accounts"] = genesis_file["app_state"]["auth"]["accounts"] + account_info

    num_accounts_created = len([account for account in account_info if account != 0])
    print(f'Created {num_accounts_created} accounts')

    assert num_accounts_created == number_of_accounts
    write_genesis_file(genesis_file_path, genesis_file)


if __name__ == "__main__":
    main()
