import { JSX, ParentProps } from "solid-js";

interface Props extends ParentProps {
	onclick?: JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
	class?: string;
	tabindex?: number;
}

export default function ActionButton(props: Props) {
	return (
		<button
			type="button"
			class={`font-medium text-sm tracking-wider cursor-pointer ${props.class ?? "text-zinc-500"}`}
			onclick={props.onclick}
			tabindex={props.tabindex}
		>
			{props.children}
		</button>
	);
}
