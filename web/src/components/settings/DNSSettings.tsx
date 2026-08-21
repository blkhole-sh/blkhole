import TextInput from "../form/TextInput";

interface Props {
	upstreamDNS: string;
	onUpstreamDNSInput: (value: string) => void;
	error?: string | undefined;
}

export default function DNSSettings(props: Props) {
	return (
		<section class="flex flex-col gap-6">
			<h2 class="font-medium text-sm tracking-wider text-zinc-700">
				DNS CONFIGURATION
			</h2>
			<TextInput
				label="UPSTREAM DNS SERVER"
				value={props.upstreamDNS}
				onInput={(event) => props.onUpstreamDNSInput(event.currentTarget.value)}
				error={props.error}
			/>
		</section>
	);
}
