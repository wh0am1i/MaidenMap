import { Button } from "@/components/ui/button";

export default function App() {
  return (
    <div className="min-h-screen p-8">
      <h1 className="text-2xl font-semibold">MaidenMap</h1>
      <Button className="mt-4" onClick={() => alert("click")}>
        Test button
      </Button>
    </div>
  );
}
