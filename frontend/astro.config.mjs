// @ts-check
import { defineConfig, envField } from "astro/config";
import react from "@astrojs/react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  integrations: [react()],
  vite: {
    optimizeDeps: {
      include: ["react-dom/client"],
    },
    plugins: [tailwindcss()],
  },
  env: {
    schema: {
      PUBLIC_API_URL: envField.string({
        context: "client",
        access: "public",
        default: "http://localhost:3000",
      }),
    },
  },
});
