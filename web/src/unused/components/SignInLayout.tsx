import { ParentProps } from "solid-js";

export default function SignInLayout(props: ParentProps) {
	return (
		<div class="min-h-screen flex items-center justify-center bg-white">
			{props.children}
		</div>
	);
}
