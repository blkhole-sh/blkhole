import { type JSX, type ParentProps, Show } from "solid-js";
import ButtonSolid from "../ui/ButtonSolid";
import Divider from "../ui/Divider";

interface Props extends ParentProps {
	title: string;
	description: string;
	cta?: string;
	onCTA?: () => void;
	actions?: JSX.Element;
}

export default function PageShell(props: Props) {
	return (
		<div class="px-24 py-4 flex flex-col flex-1">
			<div class="flex flex-row justify-between items-end">
				<div class="flex flex-col gap-4">
					<h1 class="font-display text-5xl tracking-tight">{props.title}</h1>
					<p class="text-zinc-500 max-w-2xl">{props.description}</p>
				</div>
				<Show when={props.cta}>
					<ButtonSolid onclick={() => props.onCTA?.()}>{props.cta}</ButtonSolid>
				</Show>
				<Show when={props.actions}>{props.actions}</Show>
			</div>
			<Divider class="mt-12" />
			{props.children}
			<Divider class="mt-auto" />
		</div>
	);
}
