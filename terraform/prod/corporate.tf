resource "aws_security_group" "corporate-ssh" {
  vpc_id      = "${aws_vpc.prod.id}"
  name        = "corporate-ssh"
  description = "Allow SSH traffic from the Bastion host"

  ingress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [
      "${aws_security_group.bastion.id}"]
  }

  egress {
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [
      "${aws_security_group.bastion.id}"]
  }

  tags {
    Name = "SSH from Bastion to Corporate"
  }
}

resource "aws_security_group" "corporate-elb-inet" {
  vpc_id      = "${aws_vpc.prod.id}"
  name        = "corporate-elb-inet"
  description = "Allow Internet traffic to and from ELB"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  egress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  tags {
    Name = "Internet to ELB Corporate"
  }
}

resource "aws_security_group" "corporate-elb-ec2" {
  vpc_id      = "${aws_vpc.prod.id}"
  name        = "corporate-elb-ec2"
  description = "Allow Traffic from ELB to EC2"

  ingress {
    from_port = 80
    to_port   = 80
    protocol  = "tcp"
    self      = true
  }

  egress {
    from_port       = 80
    to_port         = 80
    protocol        = "tcp"
    security_groups = [
      "${aws_security_group.corporate-elb-vpc.id}"]
  }

  tags {
    Name = "Allow Traffic from ELB to EC2"
  }
}

resource "aws_security_group" "corporate-elb-vpc" {
  vpc_id      = "${aws_vpc.prod.id}"
  name        = "corporate-elb"
  description = "Allow EC2 traffic to and from the ELB"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  egress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = [
      "0.0.0.0/0"]
  }

  tags {
    Name = "ELB Frontend to EC2"
  }
}

# ELB
resource "aws_elb" "corporate" {
  name                        = "corporate"
  cross_zone_load_balancing   = true
  idle_timeout                = 300
  connection_draining         = true
  connection_draining_timeout = 10
  subnets                     = [
    "${aws_subnet.public-a.id}",
    "${aws_subnet.public-b.id}"]
  security_groups             = [
    "${aws_security_group.corporate-elb-inet.id}",
    "${aws_security_group.corporate-elb-ec2.id}"]

  listener {
    lb_port           = 80
    lb_protocol       = "http"

    instance_port     = 8083
    instance_protocol = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 2
    target              = "TCP:80"
    interval            = 5
  }

  tags {
    Name = "Corporate"
  }
}
