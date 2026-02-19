import { JSX, ParentProps } from "solid-js";

interface Props extends ParentProps {
	onclick?: JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
	class?: string;
}

export default function ActionButton(props: Props) {
	return (
		<button
			type="button"
			class={`font-inter font-medium text-sm tracking-wider cursor-pointer ${props.class ?? "text-zinc-500"}`}
			onclick={props.onclick}
		>
			{props.children}
		</button>
	);
}
