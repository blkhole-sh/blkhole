import { createResource, Show } from "solid-js";
import { getQuote } from "~/lib/api";

export default function Blocked() {
	const [quote] = createResource(getQuote);

	return (
		<Show when={quote()}>
			{(quote) => (
				<div class="font-hedvig w-full h-[100vh] flex justify-center items-center">
					<div class="w-9/12 flex flex-col items-center">
						<p class="mb-8 text-base text-center">
							This page has been blocked.
						</p>
						<p class="mb-4 text-2xl text-center">“{quote().quote}”</p>
						<p class="mb-8 text-lg text-center">~{quote().author}</p>
					</div>
				</div>
			)}
		</Show>
	);
}
