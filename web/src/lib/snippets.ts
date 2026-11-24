export interface SnippetInput {
  endpoint: string;
  region: string;
  accessKeyId: string;
  secretAccessKey?: string;
  bucket?: string;
  profile: string;
}

export interface Snippets {
  env: string;
  awsCli: string;
  rclone: string;
}

export function connectionSnippets(input: SnippetInput): Snippets {
  const endpoint = input.endpoint || "<s3-endpoint>";
  const secret = input.secretAccessKey || "<secret-access-key>";
  const profile = profileName(input.profile);
  const target = input.bucket ? ` s3://${input.bucket}` : "";

  return {
    env: [
      `export AWS_ACCESS_KEY_ID=${input.accessKeyId}`,
      `export AWS_SECRET_ACCESS_KEY=${secret}`,
      `export AWS_DEFAULT_REGION=${input.region}`,
      `export AWS_ENDPOINT_URL=${endpoint}`,
    ].join("\n"),
    awsCli: [
      "# ~/.aws/config",
      `[profile ${profile}]`,
      `region = ${input.region}`,
      `endpoint_url = ${endpoint}`,
      "",
      "# ~/.aws/credentials",
      `[${profile}]`,
      `aws_access_key_id = ${input.accessKeyId}`,
      `aws_secret_access_key = ${secret}`,
      "",
      `aws --profile ${profile} s3 ls${target}`,
      "",
    ].join("\n"),
    rclone: [
      `[${profile}]`,
      "type = s3",
      "provider = Other",
      `access_key_id = ${input.accessKeyId}`,
      `secret_access_key = ${secret}`,
      `region = ${input.region}`,
      `endpoint = ${endpoint}`,
      "force_path_style = true",
    ].join("\n"),
  };
}

function profileName(name: string): string {
  const cleaned = name
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return cleaned || "garage";
}
