import { JoinBoardClient } from "./_components/JoinBoardClient";

export async function generateStaticParams() {
  return [{ token: "placeholder" }];
}

export default async function JoinBoardPage({
  params,
}: {
  params: Promise<{ token: string }>;
}) {
  const { token } = await params;
  return <JoinBoardClient token={token} />;
}
