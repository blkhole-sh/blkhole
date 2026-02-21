import { JSX, ParentProps } from "solid-js";
import { cx } from "~/lib/utils";

interface Props extends ParentProps {
	onclick?: JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
	class?: string;
	type?: "button" | "submit" | "reset";
}

export default function ButtonGhost(props: Props) {
	return (
		<button
			type={props.type ?? "button"}
			class={cx(
				"px-4 font-medium text-zinc-500 cursor-pointer",
				props.class || "py-2 text-sm",
			)}
			onclick={props.onclick}
		>
			{props.children}
		</button>
	);
}
