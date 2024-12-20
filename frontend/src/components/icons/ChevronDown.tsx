import { mergeProps } from "solid-js";

interface Props {
	class?: string;
}

export default function ChevronDown(props: Props) {
	const merged = mergeProps({ class: "size-4 text-gray-500" }, props);

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
				d="m19.5 8.25-7.5 7.5-7.5-7.5"
			/>
		</svg>
	);
}
