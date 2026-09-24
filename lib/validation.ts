export function normalizeXHandle(value: string): string {
  return value.trim().replace(/^@/, "");
}

export function validateXHandle(value: string): boolean {
  return /^[A-Za-z0-9_]{1,15}$/.test(normalizeXHandle(value));
}

export function isEthereumAddress(value: string): boolean {
  return /^0x[a-fA-F0-9]{40}$/.test(value.trim());
}
