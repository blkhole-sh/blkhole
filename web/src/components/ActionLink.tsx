import { ParentProps } from "solid-js";

interface Props extends ParentProps {
	href: string;
	download?: string;
	rel?: string;
	class?: string;
}

export default function ActionLink(props: Props) {
	return (
		<a
			href={props.href}
			download={props.download}
			rel={props.rel}
			class={`font-inter font-medium text-sm tracking-wider cursor-pointer ${props.class ?? "text-zinc-500"}`}
		>
			{props.children}
		</a>
	);
}
