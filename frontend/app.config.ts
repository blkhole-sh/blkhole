import { defineConfig } from "@solidjs/start/config";

export default defineConfig({
  server: {
    port: process.env.APP_PORT ? parseInt(process.env.APP_PORT) : 3000
  }
});
