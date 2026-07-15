import { BoardDetailClient } from "./_components/BoardDetailClient";

export async function generateStaticParams() {
  return [{ id: "placeholder" }];
}

export default async function BoardDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <BoardDetailClient id={id} />;
}
