import { defineConfig } from "orval";

export default defineConfig({
  api: {
    input: {
      target: "../shared/api/v1/openapi.yaml",
    },
    output: {
      target: "./src/api/generated",
      client: "axios",
      mode: "tags",
      clean: true,
      override: {
        mutator: {
          path: "./src/api/mutator.ts",
          name: "customInstance",
        },
      },
    },
  },
});
