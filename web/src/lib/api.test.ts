import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, api } from "./api";

function mockFetch(status: number, body: unknown) {
  const fn = vi.fn().mockImplementation(
    async () =>
      new Response(body === undefined ? null : JSON.stringify(body), {
        status,
        headers: { "Content-Type": "application/json" },
      }),
  );
  vi.stubGlobal("fetch", fn);
  return fn;
}

afterEach(() => vi.unstubAllGlobals());

describe("api", () => {
  it("calls GET operations with encoded query parameters", async () => {
    const fetch = mockFetch(200, { id: "b1" });
    await api.getBucketInfo("a b");
    expect(fetch).toHaveBeenCalledWith("/api/v2/GetBucketInfo?id=a+b", expect.objectContaining({ method: "GET" }));
  });

  it("sends POST bodies as JSON", async () => {
    const fetch = mockFetch(200, {});
    await api.allowBucketKey("bucket", "GK1", { read: true, write: false, owner: false });
    const [url, init] = fetch.mock.calls[0];
    expect(url).toBe("/api/v2/AllowBucketKey");
    expect(init.method).toBe("POST");
    expect(JSON.parse(init.body)).toEqual({
      bucketId: "bucket",
      accessKeyId: "GK1",
      permissions: { read: true, write: false, owner: false },
    });
  });

  it("requests the secret key only when asked", async () => {
    const fetch = mockFetch(200, {});
    await api.getKeyInfo("GK1");
    await api.getKeyInfo("GK1", true);
    expect(fetch.mock.calls[0][0]).toBe("/api/v2/GetKeyInfo?id=GK1");
    expect(fetch.mock.calls[1][0]).toBe("/api/v2/GetKeyInfo?id=GK1&showSecretKey=true");
  });

  it("loads the ui config from the backend", async () => {
    const fetch = mockFetch(200, { s3Endpoint: "https://s3.example.test", s3Region: "garage" });
    await expect(api.uiConfig()).resolves.toEqual({ s3Endpoint: "https://s3.example.test", s3Region: "garage" });
    expect(fetch.mock.calls[0][0]).toBe("/api/config");
  });

  it("tolerates empty success bodies", async () => {
    mockFetch(200, undefined);
    await expect(api.deleteBucket("b1")).resolves.toBeUndefined();
  });

  it("raises ApiError with Garage's message", async () => {
    mockFetch(400, {
      code: "BucketNotEmpty",
      message: "Bucket b1 is not empty",
      region: "garage",
      path: "/v2/DeleteBucket",
    });
    const error = await api.deleteBucket("b1").catch((e) => e);
    expect(error).toBeInstanceOf(ApiError);
    expect(error.status).toBe(400);
    expect(error.message).toBe("Bucket b1 is not empty");
  });

  it("falls back to the HTTP status when the error body is not JSON", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response("upstream down", { status: 502 })));
    const error = await api.listBuckets().catch((e) => e);
    expect(error).toBeInstanceOf(ApiError);
    expect(error.message).toBe("HTTP 502");
  });
});
