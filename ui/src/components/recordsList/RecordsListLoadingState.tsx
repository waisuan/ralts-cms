export default function RecordsListLoadingState() {
  return (
    <div className="space-y-6">
      <div className="flex justify-center items-center py-12">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading machines...</p>
        </div>
      </div>
    </div>
  );
}
