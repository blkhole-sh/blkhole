import { Show } from "solid-js";

export default function Footer() {
	const meta = document
		.querySelector('meta[name="blkhole-revision"]')
		?.getAttribute("content");
	const hash = meta && meta !== "__BLKHOLE_REVISION__" ? meta : null;

	return (
		<div class="px-24 py-8 flex flex-row justify-end text-sm text-zinc-500 tracking-wider">
			<Show when={hash}>
				<span>{hash?.toUpperCase()}</span>
			</Show>
		</div>
	);
}
