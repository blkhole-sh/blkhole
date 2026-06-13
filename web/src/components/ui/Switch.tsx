interface Props {
	checked?: boolean;
	onChange?: (checked: boolean) => void;
}

export default function Switch(props: Props) {
	const toggle = () => {
		props.onChange?.(!props.checked);
	};

	return (
		<button
			type="button"
			role="switch"
			aria-checked={props.checked ?? false}
			onclick={toggle}
			class="inline-flex h-6 w-10 shrink-0 cursor-pointer items-center px-1 transition-colors"
			classList={{ "bg-black": props.checked, "bg-zinc-200": !props.checked }}
		>
			<span
				class="bg-white pointer-events-none block size-4 shrink-0 shadow transition-transform"
				classList={{ "translate-x-4": props.checked, "translate-x-0": !props.checked }}
			/>
		</button>
	);
}
