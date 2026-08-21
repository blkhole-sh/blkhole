import type { JSX, ParentProps } from "solid-js";
import { cx } from "~/lib/utils";

interface Props extends ParentProps {
	onclick?: JSX.EventHandlerUnion<HTMLButtonElement, MouseEvent>;
	class?: string;
	tabindex?: number;
	type?: "button" | "submit" | "reset";
	disabled?: boolean;
}

export default function ActionButton(props: Props) {
	return (
		<button
			type={props.type ?? "button"}
			class={cx(
				"font-medium text-sm tracking-wider",
				props.disabled ? "cursor-default" : "cursor-pointer",
				props.class || "text-zinc-700 hover:text-black",
			)}
			onclick={props.onclick}
			tabindex={props.tabindex}
			disabled={props.disabled}
		>
			{props.children}
		</button>
	);
}
