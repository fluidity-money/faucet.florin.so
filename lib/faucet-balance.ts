// Low-ETH warning for the faucet page.
//
// Claims are paid by the sender wallets whitelisted in src/Faucet.sol
// (senderOk), so "the faucet's ETH" is the sum of their balances. All of them
// are read in one eth_call through Multicall3's getEthBalance, because the
// public RPC rate-limits a JSON-RPC batch of 100 getBalance calls.

const DEFAULT_RPC_URL = "https://rpc.testnet.chain.robinhood.com";
const MULTICALL3 = "ca11bde05977b3631167028862be2a173976ca11";
const AGGREGATE3 = "82ad56cb";
const GET_ETH_BALANCE = "4d2301cc";

// Below this total the page shows the out-of-ETH banner.
export const LOW_ETH_WEI = 5n * 10n ** 16n; // 0.05 ETH

// Keep in sync with senderOk in src/Faucet.sol.
export const SENDER_WALLETS = [
  "0xFFe12cD2Ec51E9cb832ffc55627A79D7e4203Ef5",
  "0xbd959435FB6c6C281dB6Ce6aF77bc1A802178655",
  "0x2cD955D6dE7F7160ED95922d56CC0EA53BA5EFb2",
  "0x841c86827977fc40168b4a4185b5052fcA7386c4",
  "0x44acb43890e075fcEd99C00B5C2b260273a66C0E",
  "0x983FE4F0ba0fAbDb4D541AC9DE1B9feFC46EB5F4",
  "0xeC2a1B8bf0508088ad4b022A679B666D31339a9f",
  "0x7F1Fd782044D152ddD97b1f2942cCA54E35D64Da",
  "0xcc4b50d419B6917d3fDACb9D5EA06da7DdBfe145",
  "0x0050DcDcd3959132F1688Dc95d0488cCF0Bd2515",
  "0xa06815A212c95f4A6F09aC3f9Fd74B4518e73823",
  "0xE2C1C265146e4a61f1b2899f01a95A36cB11Cd9b",
  "0x0604c23D08017bc8Ee3d465870DE93879B346cE1",
  "0x0EE31F66905B22bf5Feb75fD46f7ae5584DDBbD0",
  "0xb563c9649CcEF36d2C4fB79a85F64eDb508662d3",
  "0x32A09750F978AF48600A09A3240093fACadA32Fd",
  "0xD5959f10c789a545b79BDcf2e670FB522aBF3e7f",
  "0xd6249481D629aD8638F06c2D0C9D9FB52211B7C5",
  "0x348b3BC0eb82598FEBd7ca81230172Bb41B72DF9",
  "0xb9048B00ceAaF199576A26ED181C2E8Df118021a",
  "0x03543442e442cD74b681Aa7C8185B47056d9B2e1",
  "0x9239c390da13c8dC0263eC985f6221a28924A581",
  "0xB263DDCd1eD1Ce1335F9909b823797c2531eE46b",
  "0x4AffD54F0809fc4185b870e50B08E082948Af5E4",
  "0x3e065c97F5892Ab9D4F3E114819A14ed7A11d76d",
  "0xEb6945BE44c11A9E821a91A5728167B1125B08cA",
  "0x0019Ef523803F88E08F0e8bbF5653c6177284ae4",
  "0x834a1620b04ac4De8232445E5073407DE3FaC3f6",
  "0x24060B7aD5023c474eC2D668A0aacB4442e9E044",
  "0x53E3DC5651b54f07038c115DC4C14C3f1c162659",
  "0xd1AC368b02D332Ad72CA4B96f578b0d896717d9B",
  "0x20F95AF9CD5A440345620C842820bB9fa2667ABF",
  "0x5fb93A9671d578bCe2d8e3992F258B35Cf5b515c",
  "0x0E7E61826dDe37429E343E0000172261b2622e12",
  "0xFC246ce70C202773936Ad85Ff9F3B3f87EfE4140",
  "0x837E63b70812f6348556b36b87605CaA4987FDc5",
  "0x0d7f1aE5e937EEC7bA33ab2b9a70244eba860A0F",
  "0xb637C3c01B81E1Db57d56F1C43086dC1d5F67d14",
  "0xb8Fe47904B1953717b53f3594a58E9F9Cd375847",
  "0x20Aa286da300d12114326637064789EB71e77899",
  "0x328d3B319FF47042b27539491E7A961B97DFC23A",
  "0xd4A890D1244b14fE324633AD3A7257FaF77a22f9",
  "0x7f7C174C3D39672B67252Aec349B39504f4863DA",
  "0x26B4C0363f07199CcbCe2214a7d97465089e1F81",
  "0x8AE7E92c3dcDd1d34fEB0383Dd5A751A5F8766b6",
  "0x09526E658347e6B5064ab211E88602C15C5B0644",
  "0x2db7c80D65F86bD8D050C0b68843d2812cbfe914",
  "0x478F5eE48E74AA6c23EB6990Fe9705c01F65bF4B",
  "0xd35e673bdD5453007814839A94483D49d27f1F11",
  "0x49A71Cf91190B00038E977792aBCAcE1545a04C0",
  "0xD0758e60f644D3ae43Ba83207b2f7F2F0C659815",
  "0x51F271be12019cfAe6A2299563540bb3ef3117f2",
  "0x38074e811E2691085D289907EC4D08C258161158",
  "0x46dDA652914624B1fffC19F8e5AE711d030a0bE6",
  "0xEe152839bc7E8440074Adf07c46540F508E7360c",
  "0x1641B0623144810F91829e989e73F2E0aEE6b323",
  "0x9a82191f047E60d48E63E809fdBB533d6f3bB60b",
  "0x9e671918157e72c7a3Bac16669681a5e3696f2E2",
  "0xFf52Fa16873b8fA2F7376E16DcaEaab6ACBceE66",
  "0x96b754087499bFAef9e6E537dC535Af78eA3e9be",
  "0xF9Ec86d045eF27Af7AFCAf4d36844309203c3737",
  "0x4811dA11de0ed6d4F464b683ebC2663913b65cCc",
  "0x9d13C3Ae50F7edEF27c9490F7ae0085184d2A4ee",
  "0x4ee3C1118678fd41c0FE84b8208D4880A20aF80b",
  "0x16c2DB487640a556588DB6Dd6fC36ccEBBaF319d",
  "0x8205AcF5ad17b3227E25Cd15C957fA3f787e3bbD",
  "0xcfa12b454EAB36fE269A30DaAf0A4D818546B47A",
  "0x1C48a39B06F602eF644d6C5afBfA09d7fD31a9C7",
  "0x97D57D124dFF202d8Cd1697d0b594386f6145f56",
  "0x6beF865794ffaB6305ae975Eea0E5FE084a356D7",
  "0x3ddAA4d6D9fbDEb636DEB7344d169D7FE98eb5eD",
  "0x27Ba66eFE2E3C6d1dd2495d7b707f5983Baf803f",
  "0xd70aE73c326BDC12DC245A7b6490773FE39D45A9",
  "0xF6819D47f26f755277713d595807f6D4aB660099",
  "0x0416416d5eACFe1841ad60Ca8EA4323b7Fd3F5bf",
  "0x4290C5B105Bb65E6a2D80820DF04A3d3781162fA",
  "0x93884992094Fa2F2c3Ff82173aef8D328aC66573",
  "0x1CE0eaaAffcFF192D27CC59973aA94425DD60842",
  "0xDB013B9d66FEFF4d9A0028D02361AA66D3d8dEA2",
  "0x2831f124D2394530ECb33BCC7CE550CF68caa47E",
  "0x755155EaBb494D16e068E4A7fD69982F7343D163",
  "0x2dE28ebF08C820ed944e38F32b032F27d96f4D68",
  "0xd8C26f0168185f2EfE7F0415fC08a45484A0bEB5",
  "0xFa645972419687dC67c5ed70F12B130B22b7F98D",
  "0x0593eD785Da88ff8B75EBC79d6B69eefB5A1BA81",
  "0xF0f40E1bCA82730549296E0A0F6eF4a0f97d7Cb0",
  "0xc07B891de4B25B6d7f7e00fe003ef280b14893A0",
  "0x97D7cbC46ACb1859006fFaba304d5B5fF4403b44",
  "0x09236885C833CB0161Ff1B84548bAA7d30cE2722",
  "0x92c23011a094d030D9d52aeC5D48463254143214",
  "0x39c1B874e9179F3160C886558c7941B7502ea23D",
  "0x5b067BE9FEC3321aa7b2b8479fc6caF30D1a54b7",
  "0x6188699Ba73C04DDA227f25C520B6EBe9D87379c",
  "0x3312E2306deA108B66a528fEff8e87475a1357Bf",
  "0xCb1A330061e1F29E532CE31cCD48569880a9cf08",
  "0xd340772ACdA3627679739CEA22C3363679c55638",
  "0xF9Dee7E0A53497eD005Dc323138EEeaDD75d7678",
  "0xE59A0D90b66AFF0a617C29Efc35ec5bDA9149aBd",
  "0x786EFb542d874993d223db4eceb0e440e3971779",
  "0x9c24bC36fc33a9C28b7cD90eb4aBC7A530EC7899",
];

const word = (n: number | bigint) => n.toString(16).padStart(64, "0");
const addressWord = (address: string) => address.slice(2).toLowerCase().padStart(64, "0");

// aggregate3((address target, bool allowFailure, bytes callData)[]), with
// every call being Multicall3.getEthBalance(wallet).
export function encodeBalancesCall(wallets: readonly string[]): string {
  const tupleSize = 6 * 32;
  const offsets = wallets.map((_, i) => word(wallets.length * 32 + i * tupleSize));
  const tuples = wallets.map((wallet) =>
    addressWord(`0x${MULTICALL3}`) +
    word(1) +
    word(3 * 32) +
    word(36) +
    (GET_ETH_BALANCE + addressWord(wallet)).padEnd(128, "0"),
  );
  return `0x${AGGREGATE3}${word(32)}${word(wallets.length)}${offsets.join("")}${tuples.join("")}`;
}

// Sums the (bool success, bytes returnData)[] that aggregate3 returns,
// skipping any call that failed.
export function sumBalances(result: string): bigint {
  const hex = result.slice(2);
  const at = (byte: number) => BigInt(`0x${hex.slice(byte * 2, byte * 2 + 64)}`);
  const array = Number(at(0));
  const count = Number(at(array));
  let total = 0n;
  for (let i = 0; i < count; i++) {
    const tuple = array + 32 + Number(at(array + 32 + i * 32));
    if (at(tuple) === 0n) continue;
    total += at(tuple + Number(at(tuple + 32)) + 32);
  }
  return total;
}

// Total ETH across the sender wallets, or null when the RPC can't be read, in
// which case the page stays quiet rather than guessing.
export async function faucetEthBalance(): Promise<bigint | null> {
  try {
    const response = await fetch(process.env.FAUCET_RPC_URL || DEFAULT_RPC_URL, {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({
        jsonrpc: "2.0",
        id: 1,
        method: "eth_call",
        params: [{ to: `0x${MULTICALL3}`, data: encodeBalancesCall(SENDER_WALLETS) }, "latest"],
      }),
      next: { revalidate: 60 },
    });
    const body = (await response.json()) as { result?: string };
    return typeof body.result === "string" ? sumBalances(body.result) : null;
  } catch {
    return null;
  }
}
