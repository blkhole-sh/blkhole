import { ParentProps } from "solid-js";
import FormInput from "./FormInput";

interface Props {
	id: string;
	title: string;
	placeholder: string;
	rows: number;
	class: string;
	value: () => string;
	onChange: (
		e: Event & {
			target: HTMLTextAreaElement;
		},
	) => void;
}

export default function TextAreaInput(props: Props & ParentProps) {
	return (
		<FormInput id={props.id} title={props.title} class={props.class}>
			<textarea
				name={props.id}
				id={props.title}
				rows={props.rows}
				class="block w-full rounded-md border border-gray-300 bg-white px-3 py-1.5 text-base text-gray-900 placeholder:text-gray-400 sm:text-sm/6"
				placeholder={props.placeholder}
				value={props.value()}
				onChange={props.onChange}
			></textarea>
		</FormInput>
	);
}
