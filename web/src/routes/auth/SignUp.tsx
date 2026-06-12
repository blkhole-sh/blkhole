import { createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import AuthForm from "~/components/layout/AuthForm";
import TextInput from "~/components/form/TextInput";
import { useAuth } from "~/context/AuthContext";
import { compose, email as isEmail, minLength, required } from "~/lib/validate";

export default function SignUp() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);
	const [error, setError] = createSignal<string | undefined>(undefined);

	const auth = useAuth();
	const navigate = useNavigate();

	const handleSignUp = async (e: Event) => {
		e.preventDefault();
		setSubmitted(true);
		if (!email().trim() || !password() || password().length < 12) return;

		try {
			setError(undefined);
			await auth.register(email(), password());
			navigate("/");
		} catch (e) {
			setError(e instanceof Error ? e.message : "Sign up failed");
		}
	};

	return (
		<AuthForm
			onSubmit={handleSignUp}
			submitLabel="SIGN UP"
			footer={<a href="/auth/signin">Already have an account? Sign in</a>}
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
				validate={compose(required(), minLength(12))}
				showError={submitted()}
			/>
			<Show when={error()}>
				<p class="text-sm text-red-700 -mt-2 text-center">{error()}</p>
			</Show>
		</AuthForm>
	);
}
