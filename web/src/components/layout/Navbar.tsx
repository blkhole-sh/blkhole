import { type JSX, Show } from "solid-js";
import ActionButton from "../ui/ActionButton";

interface Props {
	title: string;
	description: string;
	cta?: string | undefined;
	onCTA?: (() => void) | undefined;
	actions?: JSX.Element | undefined;
}

export default function Navbar(props: Props) {
	return (
		<header class="sticky top-0 z-10 h-16 px-12 flex flex-row items-center justify-between gap-12 border-b border-zinc-200 bg-white">
			<div class="min-w-0 flex flex-row items-center gap-6">
				<h1 class="flex-shrink-0 font-display text-2xl leading-tight tracking-tight">
					{props.title}
				</h1>
				<span class="w-px h-5 flex-shrink-0 bg-zinc-200" />
				<p class="min-w-0 overflow-hidden text-ellipsis whitespace-nowrap text-sm text-zinc-500">
					{props.description}
				</p>
			</div>
			<div class="flex-shrink-0 flex flex-row items-center gap-6">
				<Show when={props.cta}>
					<ActionButton
						onclick={() => props.onCTA?.()}
						class="text-zinc-500 hover:text-black"
					>
						{props.cta}
					</ActionButton>
				</Show>
				<Show when={props.actions}>{props.actions}</Show>
			</div>
		</header>
	);
}
