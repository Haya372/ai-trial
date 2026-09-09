import { defineConfig } from "orval";

export default defineConfig({
	"ai-trial": {
		input: {
			target: "./openapi.yaml",
		},
		output: {
			target: "../../apps/web/src/api/generated.ts",
			client: "react-query",
			override: {
				query: {
					useQuery: true,
					useMutation: true,
				},
			},
		},
	},
});
