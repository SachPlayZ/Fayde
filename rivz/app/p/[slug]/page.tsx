import { PublicProjectsClient } from "./_components/PublicProjectsClient";

export async function generateStaticParams() {
  return [{ slug: "placeholder" }];
}

export default async function PublicProjectsPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  return <PublicProjectsClient slug={slug} />;
}
