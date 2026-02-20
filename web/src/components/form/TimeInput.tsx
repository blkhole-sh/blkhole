import { JSX, Show } from "solid-js";
import Divider from "../ui/Divider";

interface Props {
	label: string;
	value: string;
	onInput?: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>;
	hint?: string;
	error?: string;
	validate?: (value: string) => string | undefined;
	showError?: boolean;
	class?: string;
}

export default function TimeInput(props: Props) {
	const activeError = () =>
		props.error ??
		(props.showError ? props.validate?.(props.value) : undefined);
	const message = () => activeError() ?? props.hint;
	const isError = () => !!activeError();

	return (
		<div class={`flex flex-col gap-1 ${props.class ?? ""}`}>
			<label
				for={props.label}
				class="font-medium text-zinc-700 text-sm tracking-wider"
			>
				{props.label}
			</label>
			<input
				id={props.label}
				type="time"
				value={props.value}
				onInput={props.onInput}
				class="w-full py-2 text-sm leading-snug tracking-wider outline-none bg-transparent appearance-none"
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
