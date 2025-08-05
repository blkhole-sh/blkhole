import { ParentProps } from "solid-js";
import FormInput from "./FormInput";

interface Props {
	id: string;
	title: string;
	placeholder: string;
	class?: string;
	type?: string;
	required?: boolean;
	value: () => string;
	onChange: (
		e: Event & {
			target: HTMLInputElement;
		},
	) => void;
}

export default function TextInput(props: Props & ParentProps) {
	return (
		<FormInput id={props.id} title={props.title} class={props.class}>
			<input
				type={props.type || "text"}
				name={props.id}
				id={props.id}
				required={props.required}
				class="block w-full rounded-md border border-gray-300 py-1.5 px-3 text-base text-gray-900 placeholder:text-gray-400 sm:text-sm/6"
				placeholder={props.placeholder}
				value={props.value()}
				onChange={props.onChange}
			/>
		</FormInput>
	);
}
