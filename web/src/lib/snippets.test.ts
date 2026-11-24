import { describe, expect, it } from "vitest";
import { connectionSnippets } from "./snippets";

const input = {
  endpoint: "https://s3.example.test",
  region: "garage",
  accessKeyId: "GK123",
  secretAccessKey: "s3cr3t",
  bucket: "backups",
  profile: "ci-runner",
};

describe("connectionSnippets", () => {
  it("renders shell environment variables", () => {
    expect(connectionSnippets(input).env).toBe(
      [
        "export AWS_ACCESS_KEY_ID=GK123",
        "export AWS_SECRET_ACCESS_KEY=s3cr3t",
        "export AWS_DEFAULT_REGION=garage",
        "export AWS_ENDPOINT_URL=https://s3.example.test",
      ].join("\n"),
    );
  });

  it("renders an aws cli example against the endpoint and bucket", () => {
    const { awsCli } = connectionSnippets(input);
    expect(awsCli).toContain("[profile ci-runner]");
    expect(awsCli).toContain("endpoint_url = https://s3.example.test");
    expect(awsCli).toContain("aws --profile ci-runner s3 ls s3://backups");
  });

  it("renders an rclone remote", () => {
    expect(connectionSnippets(input).rclone).toBe(
      [
        "[ci-runner]",
        "type = s3",
        "provider = Other",
        "access_key_id = GK123",
        "secret_access_key = s3cr3t",
        "region = garage",
        "endpoint = https://s3.example.test",
        "force_path_style = true",
      ].join("\n"),
    );
  });

  it("uses placeholders for unknown values", () => {
    const snippets = connectionSnippets({ ...input, endpoint: "", secretAccessKey: undefined, bucket: undefined });
    expect(snippets.env).toContain("AWS_SECRET_ACCESS_KEY=<secret-access-key>");
    expect(snippets.env).toContain("AWS_ENDPOINT_URL=<s3-endpoint>");
    expect(snippets.awsCli).toContain("s3 ls\n");
  });

  it("sanitizes profile names", () => {
    expect(connectionSnippets({ ...input, profile: "My Key/1" }).rclone).toMatch(/^\[my-key-1\]/);
    expect(connectionSnippets({ ...input, profile: "" }).rclone).toMatch(/^\[garage\]/);
  });
});
