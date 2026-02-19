import { JSX, createSignal, Show } from "solid-js";
import Divider from "./Divider";
import Eye from "./icons/Eye";
import EyeSlash from "./icons/EyeSlash";

interface Props {
	label: string;
	value: string;
	onInput?: JSX.EventHandlerUnion<HTMLInputElement, InputEvent>;
	type?: "text" | "email" | "password";
	placeholder?: string;
	error?: string;
}

export default function TextInput(props: Props) {
	const [showPassword, setShowPassword] = createSignal(false);
	const inputType = () =>
		props.type === "password" && showPassword() ? "text" : props.type ?? "text";

	return (
		<div class="flex flex-col gap-1 font-inter">
			<label for={props.label} class="font-medium text-zinc-500 text-sm">
				{props.label}
			</label>
			<div class="relative w-full">
				<input
					id={props.label}
					type={inputType()}
					placeholder={props.placeholder ?? ""}
					value={props.value}
					onInput={props.onInput}
					class="w-full py-2 text-sm tracking-wider outline-none focus:border-black"
					classList={{ "pr-8": props.type === "password", "border-red-500": !!props.error }}
				/>
				<Show when={props.type === "password"}>
					<button
						type="button"
						onclick={() => setShowPassword((v) => !v)}
						class="absolute right-1 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600 cursor-pointer"
						aria-label={showPassword() ? "Hide password" : "Show password"}
					>
						<Show when={showPassword()} fallback={<Eye />}>
							<EyeSlash />
						</Show>
					</button>
				</Show>
			</div>
			<Divider />
			<Show when={props.error}>
				<p class="text-xs text-red-500">{props.error}</p>
			</Show>
		</div>
	);
}
