import { JSX, ParentProps } from "solid-js";
import { cx } from "~/lib/utils";

interface Props extends ParentProps {
	onclick?: JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
	class?: string;
	tabindex?: number;
	type?: "button" | "submit" | "reset";
}

export default function ActionButton(props: Props) {
	return (
		<button
			type={props.type ?? "button"}
			class={cx(
				"font-medium text-sm tracking-wider cursor-pointer",
				props.class || "text-zinc-500",
			)}
			onclick={props.onclick}
			tabindex={props.tabindex}
		>
			{props.children}
		</button>
	);
}
