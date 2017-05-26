#!/usr/bin/perl

package TSCollector;

require Exporter;

use IPC::Open2;
use JSON;

our @ISA     = qw(Exporter);
our @EXPORT  = qw();
our $VERSION = 1.00;

use warnings;
use strict;

# 0: url
# 1: username
# 2: password
# 3: data
sub post($$$$) {
    my ($url, $username, $password, $data) = @_;

    my $json = encode_json($data);

    my $output;
    my $pid = open2(\*output, undef, 'wget', '-q', '-O', '-', "--user=$username", "--password=$password", "--post-data=$json", $url);
    my $response;
    {
        local $/ = "";
        $response = <output>;
    }

    waitpid( $pid, 0 );
    my $rc = $? >> 8;

    if ($rc) {
        return "wget returned $rc"
    }

    my $resp = decode_json $response;
    if (!defined $resp) {
        return "Cannot decode response JSON"
    }

    if ($resp->{Status} > 0) {
        return $resp->{Message}
    }
    
    return undef;
}

# 0: transaction
# 1: key
# 2: value
# 3: value type
sub addValue($$$;$) {
    $_[0]->{$_[1]}{Value} = $_[2] . '';
    $_[0]->{$_[1]}{Type} = $_[3] if defined $_[3];
}

sub newTransaction() {
    my %r;

    return \%r;
}
    
1;
