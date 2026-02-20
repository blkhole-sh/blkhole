import { Show } from "solid-js";

export default function Footer() {
	const version = __APP_VERSION__;
	const hash = __APP_HASH__;

	return (
		<div class="px-24 py-8 flex flex-row justify-between text-sm text-zinc-500 tracking-wider">
			<Show when={version}>
				<span>VERSION {version}</span>
			</Show>
			<Show when={hash}>
				<span>{hash!.toUpperCase()}</span>
			</Show>
		</div>
	);
}
