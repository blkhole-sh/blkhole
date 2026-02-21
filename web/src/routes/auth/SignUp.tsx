import { createSignal } from "solid-js";
import AuthForm from "~/components/layout/AuthForm";
import TextInput from "~/components/form/TextInput";
import { compose, email as isEmail, minLength, required } from "~/lib/validate";

export default function SignUp() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [submitted, setSubmitted] = createSignal(false);

	const handleSignUp = (e: Event) => {
		e.preventDefault();
		setSubmitted(true);
		if (!email().trim() || !password() || password().length < 8) return;
		// TODO: submit
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
				validate={compose(required(), minLength(8))}
				showError={submitted()}
			/>
		</AuthForm>
	);
}
