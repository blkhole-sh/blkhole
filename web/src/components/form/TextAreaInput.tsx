import { JSX, Show } from "solid-js";
import Divider from "../ui/Divider";

interface Props {
	label: string;
	value: string;
	onInput?: JSX.EventHandlerUnion<HTMLTextAreaElement, InputEvent>;
	placeholder?: string;
	hint?: string;
	error?: string;
	validate?: (value: string) => string | undefined;
	showError?: boolean;
	rows?: number;
	class?: string;
}

export default function TextAreaInput(props: Props) {
	const activeError = () =>
		props.error ??
		(props.showError ? props.validate?.(props.value) : undefined);
	const message = () => activeError() ?? props.hint;
	const isError = () => !!activeError();

	return (
		<div class="flex flex-col gap-1">
			<label
				for={props.label}
				class="font-medium text-zinc-700 text-sm tracking-wider"
			>
				{props.label}
			</label>
			<textarea
				id={props.label}
				placeholder={props.placeholder ?? ""}
				value={props.value}
				onInput={props.onInput}
				rows={props.rows}
				class={`w-full py-2 text-sm leading-snug tracking-wider outline-none resize-y ${props.class ?? ""}`}
			/>
			<Divider class={isError() ? "border-red-600" : ""} />
			<Show when={message()}>
				<p class={`text-xs ${isError() ? "text-red-700" : "text-zinc-400"}`}>
					{message()}
				</p>
			</Show>
		</div>
	);
}
