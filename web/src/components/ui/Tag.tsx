import type { ParentProps } from "solid-js";

interface TagProps extends ParentProps {
	onclick?: () => void;
}

export default function Tag(props: TagProps) {
	return (
		<div
			class="px-2 py-1 font-medium bg-zinc-100 text-black text-sm"
			classList={{ "cursor-pointer": !!props.onclick }}
			onclick={props.onclick}
		>
			{props.children}
		</div>
	);
}
