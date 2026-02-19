import { createSignal } from "solid-js";

interface Props {
	checked?: boolean;
	onChange?: (checked: boolean) => void;
}

export default function Switch(props: Props) {
	const [checked, setChecked] = createSignal(props.checked ?? false);

	const toggle = () => {
		const next = !checked();
		setChecked(next);
		props.onChange?.(next);
	};

	return (
		<button
			type="button"
			role="switch"
			aria-checked={checked()}
			onclick={toggle}
			class="inline-flex h-6 w-10 shrink-0 cursor-pointer items-center rounded-full px-1 transition-colors"
			classList={{ "bg-black": checked(), "bg-zinc-200": !checked() }}
		>
			<span
				class="bg-white pointer-events-none block size-4 shrink-0 rounded-full shadow transition-transform"
				classList={{ "translate-x-4": checked(), "translate-x-0": !checked() }}
			/>
		</button>
	);
}
