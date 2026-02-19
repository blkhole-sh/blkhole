import { createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import { useAuth } from "~/context/AuthContext";
import Logo from "~/components/Logo";
import TextInput from "~/components/TextInput";
import ButtonSolid from "~/components/ButtonSolid";

export default function SignIn() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [error, setError] = createSignal<string | undefined>(undefined);

	const auth = useAuth();
	const navigate = useNavigate();

	const handleSignIn = async () => {
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
					<Show when={error()}>
						<p class="font-inter text-sm text-red-500 mb-2">{error()}</p>
					</Show>
					<ButtonSolid
						class="w-full my-1 py-4 text-base tracking-wider"
						onclick={handleSignIn}
					>
						SIGN IN
					</ButtonSolid>
					<div class="h-16 flex flex-col gap-2 mt-4 font-inter text-sm text-zinc-500 items-center">
						<a href="/auth/forgot-password">Forgot password?</a>
						<a href="/auth/signup">Create an account</a>
					</div>
				</div>
			</div>
		</main>
	);
}
