import { ParentProps } from "solid-js";

export default function Tag(props: ParentProps) {
	return (
		<div class="px-2 py-1 font-inter font-medium bg-zinc-100 text-black text-sm">
			{props.children}
		</div>
	);
}
