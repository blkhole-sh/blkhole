import { createSignal, createUniqueId, Show } from "solid-js";

interface Props {
	label: string;
	accept?: string;
	hint?: string;
	error?: string;
	showError?: boolean;
	onChange: (content: string, fileName: string) => void;
}

export default function FileInput(props: Props) {
	const [dragOver, setDragOver] = createSignal(false);
	const [fileName, setFileName] = createSignal("");
	const inputId = createUniqueId();

	const handleFile = async (file: File) => {
		const text = await file.text();
		setFileName(file.name);
		props.onChange(text, file.name);
	};

	const handleDrop = (e: DragEvent) => {
		e.preventDefault();
		setDragOver(false);
		const file = e.dataTransfer?.files[0];
		if (file) handleFile(file);
	};

	const activeError = () =>
		props.error ?? (props.showError && !fileName() ? "Required" : undefined);
	const message = () => activeError() ?? props.hint;
	const isError = () => !!activeError();

	return (
		<div class="flex flex-col gap-2">
			<p class="font-medium text-zinc-500 text-sm tracking-wider">
				{props.label}
			</p>
			<label
				for={inputId}
				class="border-2 border-dashed p-8 flex flex-col items-center gap-2 cursor-pointer"
				classList={{
					"border-black": dragOver(),
					"border-zinc-200": !dragOver(),
				}}
				onDragOver={(e) => {
					e.preventDefault();
					setDragOver(true);
				}}
				onDragLeave={() => setDragOver(false)}
				onDrop={handleDrop}
			>
				<Show
					when={fileName()}
					fallback={
						<p class="text-sm text-zinc-400">Drop a file or click to browse</p>
					}
				>
					<p class="text-sm text-black">{fileName()}</p>
					<p class="text-xs text-zinc-400">Click to replace</p>
				</Show>
				<input
					id={inputId}
					type="file"
					accept={props.accept}
					style="display: none;"
					onChange={(e) => {
						const file = e.currentTarget.files?.[0];
						if (file) handleFile(file);
					}}
				/>
			</label>
			<Show when={message()}>
				<p class={`text-xs ${isError() ? "text-red-700" : "text-zinc-400"}`}>
					{message()}
				</p>
			</Show>
		</div>
	);
}
