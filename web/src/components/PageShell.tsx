import { ParentProps, Show } from "solid-js";
import ButtonSolid from "./ButtonSolid";
import Divider from "./Divider";

interface Props extends ParentProps {
	title: string;
	description: string;
	cta?: string;
	onCTA?: () => void;
}

export default function PageShell(props: Props) {
	return (
		<div class="px-24 py-8 flex flex-col flex-1">
			<div class="flex flex-col gap-4">
				<h1 class="font-hedvig text-5xl tracking-tight">{props.title}</h1>
				<div class="flex flex-row justify-between items-baseline">
					<p class="font-inter text-zinc-500 max-w-2xl">{props.description}</p>
					<Show when={props.cta}>
						<ButtonSolid onclick={() => props.onCTA?.()}>
							{props.cta}
						</ButtonSolid>
					</Show>
				</div>
			</div>
			<Divider class="mt-12" />
			{props.children}
			<Divider class="mt-auto" />
		</div>
	);
}
