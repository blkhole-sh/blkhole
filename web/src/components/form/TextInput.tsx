import { createSignal, type JSX, Show } from "solid-js";
import Eye from "../icons/Eye";
import EyeSlash from "../icons/EyeSlash";
import Divider from "../ui/Divider";

interface Props {
	label: string;
	value: string;
	onInput?: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>;
	type?: "text" | "email" | "password";
	placeholder?: string;
	hint?: string;
	error?: string;
	validate?: (value: string) => string | undefined;
	showError?: boolean;
}

export default function TextInput(props: Props) {
	const [showPassword, setShowPassword] = createSignal(false);

	const activeError = () =>
		props.error ??
		(props.showError ? props.validate?.(props.value) : undefined);
	const message = () => activeError() ?? props.hint;
	const isError = () => !!activeError();

	return (
		<div class="flex flex-col gap-1">
			<label
				for={props.label}
				class="font-medium text-zinc-500 text-sm tracking-wider"
			>
				{props.label}
			</label>
			<div class="relative w-full">
				<Show
					when={props.type === "password" && showPassword()}
					fallback={
						<input
							id={props.label}
							type={props.type ?? "text"}
							placeholder={props.placeholder ?? ""}
							value={props.value}
							onInput={props.onInput}
							class="w-full py-2 text-sm leading-snug tracking-wider outline-none"
							classList={{ "pr-8": props.type === "password" }}
						/>
					}
				>
					<input
						id={props.label}
						type="text"
						placeholder={props.placeholder ?? ""}
						value={props.value}
						onInput={props.onInput}
						class="w-full py-2 text-sm leading-snug tracking-wider outline-none pr-8"
					/>
				</Show>
				<Show when={props.type === "password"}>
					<button
						type="button"
						onclick={() => setShowPassword((v) => !v)}
						class="absolute right-1 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600 cursor-pointer"
						aria-label={showPassword() ? "Hide password" : "Show password"}
					>
						<Show when={showPassword()} fallback={<Eye class="size-5" />}>
							<EyeSlash class="size-5" />
						</Show>
					</button>
				</Show>
			</div>
			<Divider class={isError() ? "border-red-600" : ""} />
			<Show when={message()}>
				<p class={`text-xs ${isError() ? "text-red-700" : "text-zinc-400"}`}>
					{message()}
				</p>
			</Show>
		</div>
	);
}
