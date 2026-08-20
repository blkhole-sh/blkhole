import type { JSX, ParentProps } from "solid-js";
import Navbar from "./Navbar";

interface Props extends ParentProps {
	title: string;
	description: string;
	cta?: string;
	onCTA?: () => void;
	actions?: JSX.Element;
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
			<div class="p-12 flex flex-col flex-1">{props.children}</div>
		</>
	);
}
