import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: {
      target: "./shared/api/v1/openapi.yaml",
    },
    output: {
      target: "./src/api/generated",
      client: "fetch",
      mode: "tags",
      clean: true,
    },
  },
});
