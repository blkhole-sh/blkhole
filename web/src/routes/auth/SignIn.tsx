import { useNavigate } from "@solidjs/router";
import { createSignal, Show } from "solid-js";
import TextInput from "~/components/form/TextInput";
import AuthForm from "~/components/layout/AuthForm";
import { useAuth } from "~/context/AuthContext";
import { compose, email as isEmail, required } from "~/lib/validate";

export default function SignIn() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);
	const [error, setError] = createSignal<string | undefined>(undefined);

	const auth = useAuth();
	const navigate = useNavigate();

	const handleSignIn = async (e: Event) => {
		e.preventDefault();
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
		<AuthForm
			onSubmit={handleSignIn}
			submitLabel="SIGN IN"
			footer={
				<>
					<a href="/auth/forgot-password">Forgot password?</a>
					<a href="/auth/signup">Create an account</a>
				</>
			}
		>
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
			<Show when={error()}>
				<p class="text-sm text-red-700 -mt-2">{error()}</p>
			</Show>
		</AuthForm>
	);
}
