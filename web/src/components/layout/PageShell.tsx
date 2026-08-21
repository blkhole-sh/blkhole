import type { JSX, ParentProps } from "solid-js";
import { cx } from "~/lib/utils";
import Navbar from "./Navbar";

interface Props extends ParentProps {
	title: string;
	description: string;
	cta?: string;
	onCTA?: () => void;
	actions?: JSX.Element;
	contentClass?: string;
}

export default function PageShell(props: Props) {
	return (
		<>
			<Navbar
				title={props.title}
				description={props.description}
				cta={props.cta}
				onCTA={props.onCTA}
				actions={props.actions}
			/>
			<div
				class={cx("min-h-0 flex flex-col flex-1", props.contentClass ?? "p-12")}
			>
				{props.children}
			</div>
		</>
	);
}
