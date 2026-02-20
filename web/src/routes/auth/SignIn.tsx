import { createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { useAuth } from "~/context/AuthContext";
import Logo from "~/components/ui/Logo";
import TextInput from "~/components/form/TextInput";
import ButtonSolid from "~/components/ui/ButtonSolid";
import { compose, email as isEmail, required } from "~/lib/validate";

export default function SignIn() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);
	const [error, setError] = createSignal<string | undefined>(undefined);

	const auth = useAuth();
	const navigate = useNavigate();

	const handleSignIn = async () => {
		setSubmitted(true);
		if (!email().trim() || !password()) return;
		try {
			setError(undefined);
			await auth.login(email(), password());
			navigate("/");
		} catch (e) {
			setError(e instanceof Error ? e.message : "Sign in failed");
		}
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
							validate={required()}
							showError={submitted()}
						/>
					</div>
					<Show when={error()}>
						<p class="text-sm text-red-700 mb-2">{error()}</p>
					</Show>
					<ButtonSolid
						class="w-full my-1 py-4 text-base tracking-wider"
						onclick={handleSignIn}
					>
						SIGN IN
					</ButtonSolid>
					<div class="h-16 flex flex-col gap-2 mt-4 text-sm text-zinc-500 items-center">
						<a href="/auth/forgot-password">Forgot password?</a>
						<a href="/auth/signup">Create an account</a>
					</div>
				</div>
			</div>
		</main>
	);
}
