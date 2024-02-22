
build:
	@pnpm install

data:
	@pnpm data

lint:
	@pnpm lint

serve:
	@dc up -d directus
	@pnpm dev
