import path from "node:path";
import dotenv from "dotenv";
import type { NextConfig } from "next";

dotenv.config({ path: path.resolve(process.cwd(), "../.env") });

if (!process.env.NEXT_PUBLIC_GRAPHQL_API_URL) {
  process.env.NEXT_PUBLIC_GRAPHQL_API_URL = "http://localhost:8080/graphql";
}

const nextConfig: NextConfig = {
  reactStrictMode: true,
};

export default nextConfig;
