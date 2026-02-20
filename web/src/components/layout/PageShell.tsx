import { ParentProps, Show } from "solid-js";
import ButtonSolid from "../ui/ButtonSolid";
import Divider from "../ui/Divider";

interface Props extends ParentProps {
	title: string;
	description: string;
	cta?: string;
	onCTA?: () => void;
}

export default function PageShell(props: Props) {
	return (
		<div class="px-24 py-8 flex flex-col flex-1">
			<div class="flex flex-row justify-between items-end">
				<div class="flex flex-col gap-4">
					<h1 class="font-display text-5xl tracking-tight">{props.title}</h1>
					<p class="text-zinc-500 max-w-2xl">{props.description}</p>
				</div>
				<Show when={props.cta}>
					<ButtonSolid onclick={() => props.onCTA?.()}>{props.cta}</ButtonSolid>
				</Show>
			</div>
			<Divider class="mt-12" />
			{props.children}
			<Divider class="mt-auto" />
		</div>
	);
}
