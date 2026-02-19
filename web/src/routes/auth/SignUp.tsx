import { createSignal } from "solid-js";
import Logo from "~/components/Logo";
import TextInput from "~/components/TextInput";
import ButtonSolid from "~/components/ButtonSolid";

export default function SignUp() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");

	return (
		<main class="min-h-screen flex items-center justify-center">
			<div class="w-sm">
				<Logo class="h-16 block mx-auto" />
				<div class="mt-6">
					<div class="flex flex-col gap-8 py-16 tracking-wider">
						<TextInput
							label="EMAIL"
							type="email"
							value={email()}
							onInput={(e) => setEmail(e.currentTarget.value)}
						/>
						<TextInput
							label="PASSWORD"
							type="password"
							value={password()}
							onInput={(e) => setPassword(e.currentTarget.value)}
						/>
					</div>
					<ButtonSolid class="w-full my-1 py-4 text-base tracking-wider">
						SIGN UP
					</ButtonSolid>
					<div class="h-16 flex flex-col gap-2 mt-4 font-inter text-sm text-zinc-500 items-center">
						<a href="/auth/signin">Already have an account? Sign in</a>
					</div>
				</div>
			</div>
		</main>
	);
}
