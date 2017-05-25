#!/usr/bin/perl

my $dirname;
BEGIN {
    use File::Basename;
    $dirname = dirname(__FILE__);
}

use lib $dirname;
use TSCollector;
use TSConfig;
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

my $err = TSCollector::post($TSConfig::url, $TSConfig::username, $TSConfig::password, $t);

if (defined $err) {
    print "Some error occured: $err\n"
}
