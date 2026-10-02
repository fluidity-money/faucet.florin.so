export type FaucetVariables = {
  recipientAddress: string;
  turnstileToken: string;
};

type GraphQLError = { message?: string };
type FaucetResponse = {
  data?: { requestToken?: boolean } | null;
  errors?: GraphQLError[];
};

export const REQUEST_TOKEN_MUTATION = `
  mutation RequestToken(
    $recipientAddress: String!
    $turnstileToken: String!
  ) {
    requestToken(
      recipientAddress: $recipientAddress
      turnstileToken: $turnstileToken
    )
  }
`;

export function buildFaucetRequest(variables: FaucetVariables) {
  return {
    operationName: "RequestToken",
    query: REQUEST_TOKEN_MUTATION,
    variables,
  };
}

export function readFaucetResult(response: FaucetResponse):
  | { ok: true }
  | { ok: false; message: string } {
  const message = response.errors?.find((error) => error.message)?.message;
  if (message) {
    return { ok: false, message };
  }
  if (response.data?.requestToken === true) return { ok: true };
  return {
    ok: false,
    message: "The faucet did not accept this request.",
  };
}
