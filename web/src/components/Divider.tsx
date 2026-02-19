interface Props {
	class?: string;
}

export default function Divider(props: Props) {
	return <hr class={`w-full border-t border-zinc-200 ${props.class ?? ""}`} />;
}
