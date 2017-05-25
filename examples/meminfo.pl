#!/usr/bin/perl

use TSCollector;
use warnings;
use strict;

open my $fh, "/proc/meminfo" || die "cannot open /proc/meminfo: $?";

my $t = TSCollector::newTransaction();
while (<$fh>) {
    chomp;
    next if !/^([^:]+):\s*(\d+)/;
    TSCollector::addValue($t, "meminfo/$1", $2, 'IntLast');
}
close $fh;

my $err = TSCollector::post('http://localhost:8080/add', 'test', 'test', $t);

if (defined $err) {
    print "Some error occured: $err\n"
}
