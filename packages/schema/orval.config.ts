import { defineConfig } from "orval";

export default defineConfig({
	"ai-trial": {
		input: {
			target: "./dist/openapi.yaml",
		},
		output: {
			target: "../../apps/web/src/api/generated.ts",
			client: "react-query",
      baseUrl: '/api',
			override: {
				query: {
					useQuery: true,
					useMutation: true,
				},
			},
		},
	},
	// E2E は Playwright の APIRequestContext（ログイン済み Cookie を保持）でリクエストするため、
	// 生成した fetch 関数は使わず、型と URL ビルダー（getXxxUrl）だけを apps/e2e/support/api.ts から参照する。
	e2e: {
		input: {
			target: "./dist/openapi.yaml",
		},
		output: {
			target: "../../apps/e2e/support/generated/api.ts",
			client: "fetch",
			baseUrl: "/api",
		},
	},
});
