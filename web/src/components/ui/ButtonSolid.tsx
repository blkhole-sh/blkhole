import { JSX, ParentProps } from "solid-js";

interface Props extends ParentProps {
	onclick?: JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
	class?: string;
	type?: "button" | "submit" | "reset";
}

export default function ButtonSolid(props: Props) {
	return (
		<button
			type={props.type ?? "button"}
			class={`px-4 font-medium bg-black text-white cursor-pointer ${props.class ?? "py-2 text-sm"}`}
			onclick={props.onclick}
		>
			{props.children}
		</button>
	);
}
