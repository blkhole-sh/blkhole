import { createSignal } from "solid-js";
import Logo from "~/components/ui/Logo";
import TextInput from "~/components/form/TextInput";
import ButtonSolid from "~/components/ui/ButtonSolid";
import { compose, email as isEmail, minLength, required } from "~/lib/validate";

export default function SignUp() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);

	const handleSignUp = () => {
		setSubmitted(true);
		if (!email().trim() || !password() || password().length < 8) return;
		// TODO: submit
	};

	return (
		<main class="min-h-screen flex items-center justify-center">
			<div class="w-sm">
				<Logo class="h-16 block mx-auto" />
				<div class="mt-6">
					<div class="flex flex-col gap-8 py-16 tracking-wider">
						<TextInput
							label="EMAIL"
							type="email"
							placeholder="you@example.com"
							value={email()}
							onInput={(e) => setEmail(e.currentTarget.value)}
							validate={compose(required(), isEmail())}
							showError={submitted()}
						/>
						<TextInput
							label="PASSWORD"
							type="password"
							placeholder="••••••••"
							value={password()}
							onInput={(e) => setPassword(e.currentTarget.value)}
							validate={compose(required(), minLength(8))}
							showError={submitted()}
						/>
					</div>
					<ButtonSolid
						class="w-full my-1 py-4 text-base tracking-wider"
						onclick={handleSignUp}
					>
						SIGN UP
					</ButtonSolid>
					<div class="h-16 flex flex-col gap-2 mt-4 text-sm text-zinc-500 items-center">
						<a href="/auth/signin">Already have an account? Sign in</a>
					</div>
				</div>
			</div>
		</main>
	);
}
