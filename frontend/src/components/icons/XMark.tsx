import { mergeProps } from "solid-js";

interface Props {
	class?: string;
}

export default function XMark(props: Props) {
	const merged = mergeProps({ class: "size-5 text-gray-500" }, props);

	return (
		<svg
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
			stroke-width="1.5"
			stroke="currentColor"
			class={merged.class}
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M6 18 18 6M6 6l12 12"
			/>
		</svg>
	);
}
