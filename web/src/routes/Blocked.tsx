import { createResource, Show } from "solid-js";
import Logo from "~/components/ui/Logo";
import { getQuote } from "~/lib/api";

export default function Blocked() {
	const [quote] = createResource(getQuote);

	return (
		<Show when={quote()}>
			{(quote) => (
				<>
					<div class="relative min-h-screen px-24 py-8 flex items-center justify-center">
						<a href="/" class="text-black absolute top-8 left-24">
							<Logo class="h-5" />
						</a>
						<div class="w-9/12 flex flex-col items-center">
							<p class="mb-8 text-zinc-500 text-sm text-center tracking-wider">
								THIS PAGE HAS BEEN BLOCKED.
							</p>
							<p class="mb-4 font-display text-2xl text-center">
								"{quote().quote}"
							</p>
							<p class="mb-8 font-display text-zinc-500 text-lg text-center">
								~{quote().author}
							</p>
						</div>
					</div>
				</>
			)}
		</Show>
	);
}
