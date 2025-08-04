import { createSignal, Show } from "solid-js";
import { useNavigate } from "@solidjs/router";
import TextInput from "~/components/form/TextInput";
import { useAuth } from "~/context/AuthContext";

export default function Login() {
	const [email, setEmail] = createSignal("");
	const [password, setPassword] = createSignal("");
	const [loading, setLoading] = createSignal(false);

	const auth = useAuth();
	const navigate = useNavigate();

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
		setLoading(true);

		try {
			await auth.login(email(), password());
			navigate("/");
		} catch (err) {
			console.log("Login failed");
		} finally {
			setLoading(false);
		}
	};

	return (
		<div class="h-screen flex flex-col gap-6 justify-center items-center px-6">
			<div class="w-full max-w-sm">
				<form onSubmit={handleSubmit} class="flex flex-col gap-6">
					<TextInput
						id="email"
						title="Email address"
						type="email"
						placeholder="Enter your email"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						class=""
					/>

					<div class="flex flex-col gap-3">
						<TextInput
							id="password"
							title="Password"
							type="password"
							placeholder="Enter your password"
							value={password}
							onChange={(e) => setPassword(e.target.value)}
						/>
						<a
							href="#"
							class="text-sm text-blue-800 hover:text-blue-900 underline block"
						>
							Forgot password?
						</a>
					</div>

					<div class="flex flex-col gap-3">
						<button
							type="submit"
							disabled={loading()}
							class="w-full flex justify-center rounded-md bg-gray-900 px-3 py-1.5 text-sm/6 font-semibold text-white hover:bg-gray-700 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-black disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{loading() ? "Signing in..." : "Sign in"}
						</button>
						<p class="text-center text-sm/6 text-gray-500">
							Need access?{" "}
							<a href="#" class="text-blue-800 hover:text-blue-900 underline">
								Contact admin
							</a>
						</p>
					</div>
				</form>
			</div>
		</div>
	);
}
