interface Props {
	class?: string;
}

export default function LogoIcon(props: Props) {
	return (
		<svg
			class={props.class ?? ""}
			viewBox="0 0 40 40"
			fill="none"
			xmlns="http://www.w3.org/2000/svg"
		>
			<rect
				x="2.5"
				y="2.5"
				width="35"
				height="35"
				rx="17.5"
				stroke="currentColor"
				stroke-width="5"
			/>
			<rect
				x="10.1223"
				y="9.62012"
				width="20"
				height="20"
				rx="10"
				stroke="currentColor"
				stroke-width="3"
			/>
			<rect
				x="15.6223"
				y="15.1201"
				width="9"
				height="9"
				rx="4.5"
				stroke="currentColor"
				stroke-width="2"
			/>
		</svg>
	);
}
