import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/app/**/*.{ts,tsx}",
    "./src/components/**/*.{ts,tsx}",
    "./src/features/**/*.{ts,tsx}",
    "./src/lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        ink: "#102336",
        mist: "#edf4f4",
        cloud: "#f8fbfb",
        mint: "#cfe7df",
        teal: "#1f6f68",
        sand: "#f4e4bf",
        amber: "#c88922",
        rose: "#b9534f",
      },
      boxShadow: {
        panel: "0 18px 45px rgba(16, 35, 54, 0.10)",
      },
      backgroundImage: {
        "hero-gradient":
          "radial-gradient(circle at top left, rgba(207, 231, 223, 0.75), transparent 40%), radial-gradient(circle at top right, rgba(244, 228, 191, 0.65), transparent 32%), linear-gradient(180deg, #f8fbfb 0%, #edf4f4 100%)",
      },
    },
  },
  plugins: [],
};

export default config;
