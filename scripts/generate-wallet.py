from bip_utils import Bip39SeedGenerator, Bip44, Bip44Coins
import json

def generate_wallet():
    # Generate mnemonic
    mnemonic = Bip39SeedGenerator("NUSA Chain Wallet").Generate()
    
    # Generate from seed
    seed = Bip39SeedGenerator(mnemonic).Generate()
    
    # Generate NUSA chain wallet
    bip44_mst = Bip44.FromSeed(seed, Bip44Coins.ETHEREUM)
    bip44_acc = bip44_mst.Purpose().Coin().Account(0)
    bip44_change = bip44_acc.Change(Bip44Changes.CHAIN_EXT)
    bip44_addr = bip44_change.AddressIndex(0)
    
    return {
        "mnemonic": mnemonic,
        "private_key": bip44_addr.PrivateKey().Raw().ToHex(),
        "public_key": bip44_addr.PublicKey().RawCompressed().ToHex(),
        "address": bip44_addr.PublicKey().ToAddress(),
        "chain_code": bip44_addr.ChainCode().ToHex()
    }

if __name__ == "__main__":
    wallet = generate_wallet()
    print(json.dumps(wallet, indent=2))