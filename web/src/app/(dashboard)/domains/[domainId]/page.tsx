"use client";

import { useParams, useRouter } from "next/navigation";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { ArrowLeftIcon } from "lucide-react";
import api from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { StatusBadge } from "@/components/shared/status-badge";
import { CopyButton } from "@/components/shared/copy-button";
import { CodeBlock } from "@/components/shared/code-block";
import { Skeleton } from "@/components/ui/skeleton";

interface DnsRecord {
  type: string;
  name: string;
  value: string;
  status: string;
  ttl: string;
}

interface DomainDetail {
  id: string;
  name: string;
  status: string;
  dns_records: DnsRecord[];
  created_at: string;
}

export default function DomainDetailPage() {
  const params = useParams<{ domainId: string }>();
  const router = useRouter();
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["domain", params.domainId],
    queryFn: () =>
      api.get(`/domains/${params.domainId}`).then((res) => res.data),
  });

  const domain: DomainDetail | undefined = data?.data;

  const verifyMutation = useMutation({
    mutationFn: () =>
      api
        .post(`/domains/${params.domainId}/verify`)
        .then((res) => res.data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["domain", params.domainId],
      });
    },
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!domain) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" onClick={() => router.push("/domains")}>
          <ArrowLeftIcon className="mr-2 size-4" />
          Back to Domains
        </Button>
        <p className="text-muted-foreground">Domain not found.</p>
      </div>
    );
  }

  const dnsRecords = domain.dns_records ?? [];

  const exampleDns = dnsRecords
    .map(
      (r) =>
        `${r.name}\t${r.ttl || "3600"}\tIN\t${r.type}\t${r.value}`
    )
    .join("\n");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push("/domains")}
          >
            <ArrowLeftIcon className="mr-2 size-4" />
            Back
          </Button>
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-semibold">{domain.name}</h1>
            <StatusBadge status={domain.status} />
          </div>
        </div>
        <Button
          onClick={() => verifyMutation.mutate()}
          disabled={verifyMutation.isPending}
        >
          {verifyMutation.isPending ? "Verifying..." : "Verify Domain"}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>DNS Records</CardTitle>
        </CardHeader>
        <CardContent>
          {dnsRecords.length > 0 ? (
            <div className="rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Type</TableHead>
                    <TableHead>Name</TableHead>
                    <TableHead>Value</TableHead>
                    <TableHead>Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {dnsRecords.map((record, i) => (
                    <TableRow key={i}>
                      <TableCell>
                        <span className="rounded bg-muted px-2 py-0.5 text-xs font-mono font-medium">
                          {record.type}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm max-w-[200px] truncate">
                            {record.name}
                          </span>
                          <CopyButton value={record.name} />
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-sm max-w-[300px] truncate text-muted-foreground">
                            {record.value}
                          </span>
                          <CopyButton value={record.value} />
                        </div>
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={record.status} />
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground py-4 text-center">
              No DNS records available.
            </p>
          )}
        </CardContent>
      </Card>

      {dnsRecords.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Example DNS Configuration</CardTitle>
          </CardHeader>
          <CardContent>
            <CodeBlock code={exampleDns} />
          </CardContent>
        </Card>
      )}
    </div>
  );
}
