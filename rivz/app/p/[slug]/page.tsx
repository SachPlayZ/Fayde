import { PublicShowcaseClient } from "./_components/PublicShowcaseClient";

export async function generateStaticParams() {
  return [{ slug: "placeholder" }];
}

export default async function PublicShowcasePage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  return <PublicShowcaseClient slug={slug} />;
}
